/** @typedef {import('./stylelint-config-recess-order-groups.d.ts')} _StylelintGroupsTypes */
import propertyGroups from 'stylelint-config-recess-order/groups';

export default {
	customSyntax: 'postcss-html',
	plugins: ['stylelint-order'],
	rules: {
		'order/properties-order': propertyGroups.map((group) => ({
			...group,
			emptyLineBefore: 'always',
			noEmptyLineBetween: true,
		})),
	},
};
